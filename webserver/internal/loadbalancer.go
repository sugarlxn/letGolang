package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

//
// ======================
// LoadBalancer - 负载均衡器
// ======================
//

// LoadBalancer 负载均衡器，使用轮询策略分配请求到不同的实例
type LoadBalancer struct {
	clients     []TextToImageProvider
	nextIndex   uint32
	healthCheck bool
	mu          sync.RWMutex
	available   []bool // 记录每个实例是否可用
}

// NewLoadBalancer 创建新的负载均衡器
func NewLoadBalancer(clients []TextToImageProvider) *LoadBalancer {
	lb := &LoadBalancer{
		clients:     clients,
		nextIndex:   0,
		healthCheck: true,
		available:   make([]bool, len(clients)),
	}

	// 初始化所有实例为可用
	for i := range lb.available {
		lb.available[i] = true
	}

	// 启动健康检查
	if lb.healthCheck {
		go lb.startHealthCheck()
	}

	return lb
}

// GetNext 使用轮询算法获取下一个可用的客户端
func (lb *LoadBalancer) GetNext() TextToImageProvider {
	if len(lb.clients) == 0 {
		return nil
	}

	// 简单的轮询算法
	for i := 0; i < len(lb.clients); i++ {
		index := atomic.AddUint32(&lb.nextIndex, 1) % uint32(len(lb.clients))

		lb.mu.RLock()
		isAvailable := lb.available[index]
		lb.mu.RUnlock()

		if isAvailable {
			return lb.clients[index]
		}
	}

	// 如果所有实例都不可用，返回第一个（降级处理）
	log.Println("Warning: all instances are unavailable, using first one as fallback")
	return lb.clients[0]
}

// GetByStrategy 使用指定策略获取客户端（预留接口，支持扩展）
func (lb *LoadBalancer) GetByStrategy(strategy string, userID int64) TextToImageProvider {
	switch strategy {
	case "round-robin":
		return lb.GetNext()
	case "user-hash":
		// 根据用户 ID 哈希，保证同一用户总是使用同一实例（会话亲和性）
		index := int(userID) % len(lb.clients)
		lb.mu.RLock()
		isAvailable := lb.available[index]
		lb.mu.RUnlock()

		if isAvailable {
			return lb.clients[index]
		}
		return lb.GetNext() // 降级到轮询
	default:
		return lb.GetNext()
	}
}

// startHealthCheck 定期检查所有实例的健康状态
func (lb *LoadBalancer) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for i, client := range lb.clients {
			go lb.checkInstance(i, client)
		}
	}
}

// checkInstance 检查单个实例的健康状态
func (lb *LoadBalancer) checkInstance(index int, client TextToImageProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Ping(ctx)

	lb.mu.Lock()
	defer lb.mu.Unlock()

	if err != nil {
		if lb.available[index] {
			log.Printf("Instance %d became unavailable: %v", index, err)
			lb.available[index] = false
		}
	} else {
		if !lb.available[index] {
			log.Printf("Instance %d is now available", index)
			lb.available[index] = true
		}
	}
}

// GetStats 获取负载均衡器统计信息
func (lb *LoadBalancer) GetStats() map[string]interface{} {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	availableCount := 0
	for _, available := range lb.available {
		if available {
			availableCount++
		}
	}

	return map[string]interface{}{
		"total_instances":     len(lb.clients),
		"available_instances": availableCount,
		"current_index":       atomic.LoadUint32(&lb.nextIndex),
	}
}

// AddInstance 动态添加新实例（热更新）
func (lb *LoadBalancer) AddInstance(client TextToImageProvider) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.clients = append(lb.clients, client)
	lb.available = append(lb.available, true)

	log.Printf("Added new instance, total instances: %d", len(lb.clients))
}

// RemoveInstance 移除实例
func (lb *LoadBalancer) RemoveInstance(index int) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if index < 0 || index >= len(lb.clients) {
		return nil
	}

	lb.clients = append(lb.clients[:index], lb.clients[index+1:]...)
	lb.available = append(lb.available[:index], lb.available[index+1:]...)

	log.Printf("Removed instance %d, total instances: %d", index, len(lb.clients))
	return nil
}
