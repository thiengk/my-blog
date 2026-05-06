---
title: "Microservices Architecture - Hướng dẫn toàn diện"
description: "Tìm hiểu kiến trúc Microservices, Service Discovery với Eureka, API Gateway, Circuit Breaker, và Distributed Tracing với các ví dụ thực tế."
date: 2026-05-06
category: "cong-nghe"
tags: ["microservices", "spring-boot", "java", "architecture", "eureka", "api-gateway"]
draft: false
coverImage: "https://images.unsplash.com/photo-1558494949-ef010cbdcc31?w=800&h=400&fit=crop"
---

## 🎯 Mục Tiêu Chương Này

Sau khi hoàn thành bài viết này, bạn sẽ:
- ✅ Hiểu kiến trúc Microservices và khi nào nên dùng
- ✅ Implement Service Discovery với Eureka
- ✅ Setup API Gateway với Spring Cloud Gateway
- ✅ Implement Circuit Breaker với Resilience4j
- ✅ Setup Distributed Tracing với Zipkin
- ✅ Hiểu inter-service communication patterns

## 📚 Microservices Là Gì?

### Monolithic vs Microservices

```
MONOLITHIC:
┌─────────────────────────────────┐
│     Single Application          │
│  ┌──────────────────────────┐  │
│  │  User Service            │  │
│  │  Order Service           │  │
│  │  Payment Service         │  │
│  │  Notification Service    │  │
│  └──────────────────────────┘  │
│     Single Database             │
└─────────────────────────────────┘

MICROSERVICES:
┌──────────┐  ┌──────────┐  ┌──────────┐
│  User    │  │  Order   │  │ Payment  │
│ Service  │  │ Service  │  │ Service  │
│    DB    │  │    DB    │  │    DB    │
└──────────┘  └──────────┘  └──────────┘
```

### Khi Nào Dùng Microservices?

**✅ NÊN DÙNG:**
- Team lớn (10+ developers)
- App phức tạp, nhiều domains
- Cần scale từng phần riêng biệt
- Cần deploy độc lập
- Tech stack đa dạng

**❌ KHÔNG NÊN DÙNG:**
- Team nhỏ (< 5 developers)
- App đơn giản
- Startup giai đoạn đầu
- Chưa có DevOps infrastructure

---

## 🏗️ Microservices Components

### 1. Service Discovery (Eureka)

**Vấn đề:**
```
Service A cần gọi Service B
→ IP của Service B thay đổi khi restart
→ Làm sao Service A biết IP mới?
```

**Giải pháp: Service Registry**
```
                ┌─────────────┐
                │   Eureka    │
                │   Server    │
                └─────────────┘
                   ↑       ↑
        Register   │       │   Discover
                   │       │
        ┌──────────┴─┐   ┌─┴──────────┐
        │ Service A  │   │ Service B  │
        └────────────┘   └────────────┘
```

### 2. API Gateway

**Vấn đề:**
```
Client cần gọi nhiều services
→ Phải biết địa chỉ từng service
→ Authentication ở đâu?
→ Rate limiting ở đâu?
```

**Giải pháp: API Gateway**
```
┌────────┐
│ Client │
└────┬───┘
     │
     ↓
┌─────────────┐
│ API Gateway │ ← Single entry point
└─────────────┘
     │
     ├──→ User Service
     ├──→ Order Service
     └──→ Payment Service
```

### 3. Circuit Breaker

**Vấn đề:**
```
Service A gọi Service B
→ Service B down
→ Service A timeout
→ Cascade failure
```

**Giải pháp: Circuit Breaker**
```
CLOSED (Normal):
Service A → Service B ✅

OPEN (Service B down):
Service A → ⚡ Circuit Breaker → Fallback ✅

HALF-OPEN (Testing):
Service A → ⚡ Test → Service B
```

### 4. Distributed Tracing

**Vấn đề:**
```
Request đi qua nhiều services
→ Làm sao debug khi có lỗi?
→ Service nào chậm?
```

**Giải pháp: Distributed Tracing**
```
Request ID: abc123

Client → Gateway → User Service → Order Service
  1ms      5ms        50ms           200ms
                                      ↑
                                   Bottleneck!
```

---

## 🚀 Thực Hành: Build Microservices

### Architecture Overview

```
┌────────┐
│ Client │
└────┬───┘
     │
     ↓
┌─────────────────┐
│  API Gateway    │ :8080
└────┬────────────┘
     │
     ├──→ User Service    :8081
     ├──→ Order Service   :8082
     └──→ Product Service :8083
          ↑
          │
     ┌────┴────┐
     │ Eureka  │ :8761
     │ Server  │
     └─────────┘
```

---

## 📦 1. Setup Eureka Server

### Step 1: Create Eureka Server

**pom.xml:**
```xml
<dependencies>
    <dependency>
        <groupId>org.springframework.cloud</groupId>
        <artifactId>spring-cloud-starter-netflix-eureka-server</artifactId>
    </dependency>
</dependencies>

<dependencyManagement>
    <dependencies>
        <dependency>
            <groupId>org.springframework.cloud</groupId>
            <artifactId>spring-cloud-dependencies</artifactId>
            <version>2023.0.0</version>
            <type>pom</type>
            <scope>import</scope>
        </dependency>
    </dependencies>
</dependencyManagement>
```

**EurekaServerApplication.java:**
```java
package com.example.eureka;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.eureka.server.EnableEurekaServer;

@SpringBootApplication
@EnableEurekaServer
public class EurekaServerApplication {
    public static void main(String[] args) {
        SpringApplication.run(EurekaServerApplication.class, args);
    }
}
```

**application.yml:**
```yaml
server:
  port: 8761

eureka:
  client:
    register-with-eureka: false
    fetch-registry: false
  server:
    enable-self-preservation: false
```

### Step 2: Run Eureka Server

```bash
mvn spring-boot:run
```

**Access:** http://localhost:8761

---

## 🔧 2. Create User Service

### pom.xml

```xml
<dependencies>
    <!-- Spring Boot -->
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-web</artifactId>
    </dependency>
    
    <!-- Eureka Client -->
    <dependency>
        <groupId>org.springframework.cloud</groupId>
        <artifactId>spring-cloud-starter-netflix-eureka-client</artifactId>
    </dependency>
    
    <!-- Circuit Breaker -->
    <dependency>
        <groupId>org.springframework.cloud</groupId>
        <artifactId>spring-cloud-starter-circuitbreaker-resilience4j</artifactId>
    </dependency>
    
    <!-- Distributed Tracing -->
    <dependency>
        <groupId>io.micrometer</groupId>
        <artifactId>micrometer-tracing-bridge-brave</artifactId>
    </dependency>
    <dependency>
        <groupId>io.zipkin.reporter2</groupId>
        <artifactId>zipkin-reporter-brave</artifactId>
    </dependency>
</dependencies>
```

### UserServiceApplication.java

```java
package com.example.userservice;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.client.discovery.EnableDiscoveryClient;

@SpringBootApplication
@EnableDiscoveryClient
public class UserServiceApplication {
    public static void main(String[] args) {
        SpringApplication.run(UserServiceApplication.class, args);
    }
}
```

### application.yml

```yaml
server:
  port: 8081

spring:
  application:
    name: user-service
  
  # Distributed Tracing
  zipkin:
    base-url: http://localhost:9411
  sleuth:
    sampler:
      probability: 1.0

eureka:
  client:
    service-url:
      defaultZone: http://localhost:8761/eureka/
  instance:
    prefer-ip-address: true

# Circuit Breaker
resilience4j:
  circuitbreaker:
    instances:
      orderService:
        sliding-window-size: 10
        failure-rate-threshold: 50
        wait-duration-in-open-state: 10000
        permitted-number-of-calls-in-half-open-state: 3
```

### UserController.java

```java
package com.example.userservice.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestTemplate;
import io.github.resilience4j.circuitbreaker.annotation.CircuitBreaker;

@RestController
@RequestMapping("/api/users")
public class UserController {
    
    @Autowired
    private RestTemplate restTemplate;
    
    @GetMapping("/{id}")
    public User getUser(@PathVariable Long id) {
        return new User(id, "John Doe", "john@example.com");
    }
    
    @GetMapping("/{id}/orders")
    @CircuitBreaker(name = "orderService", fallbackMethod = "getOrdersFallback")
    public List<Order> getUserOrders(@PathVariable Long id) {
        // Call Order Service
        String url = "http://order-service/api/orders/user/" + id;
        return restTemplate.getForObject(url, List.class);
    }
    
    // Fallback method when Order Service is down
    public List<Order> getOrdersFallback(Long id, Exception ex) {
        return List.of(new Order(0L, "Service unavailable"));
    }
}
```

### RestTemplateConfig.java

```java
package com.example.userservice.config;

import org.springframework.cloud.client.loadbalancer.LoadBalanced;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestTemplate;

@Configuration
public class RestTemplateConfig {
    
    @Bean
    @LoadBalanced  // Enable service discovery
    public RestTemplate restTemplate() {
        return new RestTemplate();
    }
}
```

---

## 🛒 3. Create Order Service

### OrderServiceApplication.java

```java
package com.example.orderservice;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.client.discovery.EnableDiscoveryClient;

@SpringBootApplication
@EnableDiscoveryClient
public class OrderServiceApplication {
    public static void main(String[] args) {
        SpringApplication.run(OrderServiceApplication.class, args);
    }
}
```

### application.yml

```yaml
server:
  port: 8082

spring:
  application:
    name: order-service
  
  zipkin:
    base-url: http://localhost:9411
  sleuth:
    sampler:
      probability: 1.0

eureka:
  client:
    service-url:
      defaultZone: http://localhost:8761/eureka/
  instance:
    prefer-ip-address: true
```

### OrderController.java

```java
package com.example.orderservice.controller;

import org.springframework.web.bind.annotation.*;
import java.util.List;

@RestController
@RequestMapping("/api/orders")
public class OrderController {
    
    @GetMapping("/user/{userId}")
    public List<Order> getOrdersByUser(@PathVariable Long userId) {
        // Simulate database query
        return List.of(
            new Order(1L, userId, "Product A", 100.0),
            new Order(2L, userId, "Product B", 200.0)
        );
    }
    
    @PostMapping
    public Order createOrder(@RequestBody OrderRequest request) {
        // Create order logic
        return new Order(
            System.currentTimeMillis(),
            request.getUserId(),
            request.getProduct(),
            request.getAmount()
        );
    }
}
```

---

## 🌐 4. Setup API Gateway

### pom.xml

```xml
<dependencies>
    <dependency>
        <groupId>org.springframework.cloud</groupId>
        <artifactId>spring-cloud-starter-gateway</artifactId>
    </dependency>
    
    <dependency>
        <groupId>org.springframework.cloud</groupId>
        <artifactId>spring-cloud-starter-netflix-eureka-client</artifactId>
    </dependency>
    
    <!-- Rate Limiting -->
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-data-redis-reactive</artifactId>
    </dependency>
</dependencies>
```

### ApiGatewayApplication.java

```java
package com.example.gateway;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.client.discovery.EnableDiscoveryClient;

@SpringBootApplication
@EnableDiscoveryClient
public class ApiGatewayApplication {
    public static void main(String[] args) {
        SpringApplication.run(ApiGatewayApplication.class, args);
    }
}
```

### application.yml

```yaml
server:
  port: 8080

spring:
  application:
    name: api-gateway
  
  cloud:
    gateway:
      routes:
        # User Service Routes
        - id: user-service
          uri: lb://user-service
          predicates:
            - Path=/api/users/**
          filters:
            - name: RequestRateLimiter
              args:
                redis-rate-limiter.replenishRate: 10
                redis-rate-limiter.burstCapacity: 20
        
        # Order Service Routes
        - id: order-service
          uri: lb://order-service
          predicates:
            - Path=/api/orders/**
          filters:
            - name: CircuitBreaker
              args:
                name: orderServiceCircuitBreaker
                fallbackUri: forward:/fallback/orders
      
      # Global CORS Configuration
      globalcors:
        cors-configurations:
          '[/**]':
            allowed-origins: "*"
            allowed-methods:
              - GET
              - POST
              - PUT
              - DELETE
            allowed-headers: "*"

eureka:
  client:
    service-url:
      defaultZone: http://localhost:8761/eureka/
```

### FallbackController.java

```java
package com.example.gateway.controller;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/fallback")
public class FallbackController {
    
    @GetMapping("/orders")
    public String ordersFallback() {
        return "Order service is currently unavailable. Please try again later.";
    }
}
```

---

## 🔍 5. Setup Distributed Tracing (Zipkin)

### Run Zipkin with Docker

```bash
docker run -d -p 9411:9411 openzipkin/zipkin
```

**Access:** http://localhost:9411

### Trace Flow

```
Request: GET /api/users/1/orders

Zipkin UI shows:
┌─────────────────────────────────────────┐
│ api-gateway          [5ms]              │
│  └─ user-service     [50ms]             │
│      └─ order-service [200ms]           │
└─────────────────────────────────────────┘
Total: 255ms

→ Identify bottleneck: order-service
```

---

## 🐳 6. Docker Compose Setup

### docker-compose.yml

```yaml
version: '3.8'

services:
  # Eureka Server
  eureka-server:
    build: ./eureka-server
    ports:
      - "8761:8761"
    networks:
      - microservices
  
  # User Service
  user-service:
    build: ./user-service
    ports:
      - "8081:8081"
    environment:
      - EUREKA_CLIENT_SERVICEURL_DEFAULTZONE=http://eureka-server:8761/eureka/
      - SPRING_ZIPKIN_BASE_URL=http://zipkin:9411
    depends_on:
      - eureka-server
      - zipkin
    networks:
      - microservices
  
  # Order Service
  order-service:
    build: ./order-service
    ports:
      - "8082:8082"
    environment:
      - EUREKA_CLIENT_SERVICEURL_DEFAULTZONE=http://eureka-server:8761/eureka/
      - SPRING_ZIPKIN_BASE_URL=http://zipkin:9411
    depends_on:
      - eureka-server
      - zipkin
    networks:
      - microservices
  
  # API Gateway
  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    environment:
      - EUREKA_CLIENT_SERVICEURL_DEFAULTZONE=http://eureka-server:8761/eureka/
    depends_on:
      - eureka-server
      - user-service
      - order-service
    networks:
      - microservices
  
  # Zipkin
  zipkin:
    image: openzipkin/zipkin
    ports:
      - "9411:9411"
    networks:
      - microservices
  
  # Redis (for rate limiting)
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    networks:
      - microservices

networks:
  microservices:
    driver: bridge
```

### Start All Services

```bash
docker-compose up -d
```

### Check Services

```bash
# Eureka Dashboard
http://localhost:8761

# Zipkin Dashboard
http://localhost:9411

# Test API Gateway
curl http://localhost:8080/api/users/1
curl http://localhost:8080/api/orders/user/1
```

---

## 🔐 7. Inter-Service Communication

### Synchronous Communication (REST)

```java
@Service
public class UserService {
    
    @Autowired
    private RestTemplate restTemplate;
    
    public UserWithOrders getUserWithOrders(Long userId) {
        // Get user
        User user = getUser(userId);
        
        // Call Order Service synchronously
        String url = "http://order-service/api/orders/user/" + userId;
        List<Order> orders = restTemplate.getForObject(url, List.class);
        
        return new UserWithOrders(user, orders);
    }
}
```

### Asynchronous Communication (Kafka)

```java
@Service
public class OrderService {
    
    @Autowired
    private KafkaTemplate<String, OrderEvent> kafkaTemplate;
    
    public Order createOrder(OrderRequest request) {
        Order order = saveOrder(request);
        
        // Publish event asynchronously
        OrderEvent event = new OrderEvent(order.getId(), order.getUserId());
        kafkaTemplate.send("order-created", event);
        
        return order;
    }
}

@Service
public class NotificationService {
    
    @KafkaListener(topics = "order-created")
    public void handleOrderCreated(OrderEvent event) {
        // Send notification
        sendEmail(event.getUserId(), "Order created: " + event.getOrderId());
    }
}
```

---

## 📊 8. Monitoring Microservices

### Prometheus Configuration

**prometheus.yml:**
```yaml
scrape_configs:
  - job_name: 'user-service'
    metrics_path: '/actuator/prometheus'
    static_configs:
      - targets: ['user-service:8081']
  
  - job_name: 'order-service'
    metrics_path: '/actuator/prometheus'
    static_configs:
      - targets: ['order-service:8082']
  
  - job_name: 'api-gateway'
    metrics_path: '/actuator/prometheus'
    static_configs:
      - targets: ['api-gateway:8080']
```

### Grafana Dashboard

**Key Metrics:**
```
- Request rate per service
- Response time per service
- Error rate per service
- Circuit breaker status
- Service discovery status
```

---

## 🎯 9. Best Practices

### 1. Service Design

```
✅ DO:
- Single responsibility per service
- Independent database per service
- Stateless services
- API versioning

❌ DON'T:
- Shared database between services
- Tight coupling
- Synchronous chains (A→B→C→D)
- No fallback mechanisms
```

### 2. Communication Patterns

```
Synchronous (REST):
✅ Use for: Read operations, immediate response needed
❌ Avoid for: Long-running operations, chains

Asynchronous (Kafka):
✅ Use for: Events, notifications, long operations
❌ Avoid for: Immediate response needed
```

### 3. Error Handling

```java
@Service
public class ResilientUserService {
    
    @CircuitBreaker(name = "orderService", fallbackMethod = "fallback")
    @Retry(name = "orderService", fallbackMethod = "fallback")
    @Timeout(name = "orderService")
    public List<Order> getOrders(Long userId) {
        return orderServiceClient.getOrders(userId);
    }
    
    private List<Order> fallback(Long userId, Exception ex) {
        log.error("Failed to get orders for user: " + userId, ex);
        return Collections.emptyList();
    }
}
```

### 4. Configuration Management

```yaml
# Use Spring Cloud Config Server
spring:
  cloud:
    config:
      uri: http://config-server:8888
      fail-fast: true
      retry:
        max-attempts: 5
```

---

## 🚀 10. Deployment Strategy

### Blue-Green Deployment

```
Production:
┌─────────┐
│ Blue    │ ← 100% traffic
│ v1.0    │
└─────────┘

Deploy new version:
┌─────────┐  ┌─────────┐
│ Blue    │  │ Green   │
│ v1.0    │  │ v2.0    │
└─────────┘  └─────────┘
     ↑            ↑
   100%          0%

Switch traffic:
┌─────────┐  ┌─────────┐
│ Blue    │  │ Green   │
│ v1.0    │  │ v2.0    │
└─────────┘  └─────────┘
     ↑            ↑
    0%          100%
```

### Canary Deployment

```
Step 1: Deploy to 10%
┌─────────┐  ┌─────────┐
│ v1.0    │  │ v2.0    │
└─────────┘  └─────────┘
     ↑            ↑
    90%          10%

Step 2: Increase to 50%
     ↑            ↑
    50%          50%

Step 3: Full rollout
     ↑            ↑
    0%          100%
```

---

## 📝 Practice Exercises

### Exercise 1: Build Simple Microservices

```
Tạo 3 services:
1. Product Service (port 8081)
   - GET /api/products
   - GET /api/products/{id}

2. Inventory Service (port 8082)
   - GET /api/inventory/{productId}

3. API Gateway (port 8080)
   - Route requests to services
   - Add rate limiting

Requirements:
✅ Register with Eureka
✅ Add circuit breaker
✅ Add distributed tracing
```

### Exercise 2: Implement Circuit Breaker

```
Scenario:
- User Service calls Order Service
- Order Service randomly fails 50%
- Implement circuit breaker
- Add fallback method

Test:
1. Call endpoint 10 times
2. Observe circuit breaker opens
3. Fallback method returns cached data
```

### Exercise 3: Distributed Tracing

```
Create request flow:
Client → Gateway → User Service → Order Service → Payment Service

Requirements:
✅ Setup Zipkin
✅ Add trace IDs to logs
✅ Identify slowest service
✅ Optimize bottleneck
```

---

## 🎯 Checklist

```
✅ Understand Microservices architecture
✅ Setup Eureka Server
✅ Create multiple services
✅ Implement service discovery
✅ Setup API Gateway
✅ Add circuit breaker
✅ Setup distributed tracing
✅ Implement inter-service communication
✅ Add monitoring
✅ Deploy with Docker Compose
```

---

## 🔗 Resources

### Documentation
- [Spring Cloud](https://spring.io/projects/spring-cloud)
- [Netflix Eureka](https://github.com/Netflix/eureka)
- [Resilience4j](https://resilience4j.readme.io/)
- [Zipkin](https://zipkin.io/)

### Books
- "Building Microservices" - Sam Newman
- "Microservices Patterns" - Chris Richardson

### Courses
- Udemy: Microservices with Spring Boot
- Pluralsight: Microservices Architecture

---

## 🎉 Kết Luận

Bạn đã học:
- ✅ Kiến trúc Microservices
- ✅ Service Discovery với Eureka
- ✅ API Gateway với Spring Cloud Gateway
- ✅ Circuit Breaker với Resilience4j
- ✅ Distributed Tracing với Zipkin
- ✅ Inter-service communication patterns

**Next:** Chương tiếp theo về Kubernetes Basics

---

*"Microservices are not a silver bullet, but they solve specific problems at scale."*
