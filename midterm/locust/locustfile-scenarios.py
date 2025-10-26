"""
Resilience Patterns Load Testing with Locust

This script tests various resilience patterns implemented in the order service:
1. Normal Operation (No Failures)
2. Fail Fast Demo (Invalid Orders)
3. Circuit Breaker Demo (Inventory Failures)
4. Circuit Breaker Demo (Payment Failures)
5. Bulkhead Demo (Concurrent Requests)
6. Combined Chaos (Multiple Failures)

Usage:
    locust -f locustfile.py --host http://localhost:8080

Run with different user classes to test different scenarios:
    locust -f locustfile.py --host http://localhost:8080 --headless --users 10 --spawn-rate 2 -t 60s
"""

from locust import FastHttpUser, task, between, events, constant, LoadTestShape
import random
import time
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class ResilienceTestUser(FastHttpUser):
    """Base user class with common functionality"""
    
    # Service endpoints
    ORDER_SERVICE = ""  # Will be set from host
    INVENTORY_SERVICE = "http://localhost:8081"
    PAYMENT_SERVICE = "http://localhost:8082"
    
    # Store created order IDs for retrieval
    order_ids = []
    
    # Sample items for orders
    items = [
        {"item_id": "item-1", "quantity": 1, "price": 999.99},
        {"item_id": "item-2", "quantity": 2, "price": 29.99},
        {"item_id": "item-3", "quantity": 1, "price": 79.99},
        {"item_id": "item-4", "quantity": 1, "price": 299.99},
        {"item_id": "item-5", "quantity": 3, "price": 149.99}
    ]
    
    def create_valid_order(self, num_items=None):
        """Create a valid order with random items"""
        if num_items is None:
            num_items = random.randint(1, 3)
        
        selected_items = random.sample(self.items, min(num_items, len(self.items)))
        payload = {"items": selected_items}
        
        with self.client.post("/order/create", json=payload, catch_response=True, name="/order/create") as response:
            if response.status_code == 200:
                try:
                    order_data = response.json()
                    order_id = order_data.get("order_id")
                    if order_id and len(self.order_ids) < 100:
                        self.order_ids.append(order_id)
                    response.success()
                except Exception as e:
                    response.failure(f"Failed to parse response: {e}")
            elif response.status_code >= 500:
                # Expected during chaos scenarios
                response.success()
            else:
                response.failure(f"Got status {response.status_code}")
    
    def get_order_details(self):
        """Retrieve order details"""
        if self.order_ids:
            order_id = random.choice(self.order_ids)
            self.client.get(f"/order/{order_id}", name="/order/:orderId")
        else:
            # Try a dummy ID to test 404 handling
            self.client.get("/order/dummy-order-id", name="/order/:orderId")
    
    def check_circuit_status(self):
        """Check circuit breaker status"""
        with self.client.get("/order/circuit-status", catch_response=True, name="/order/circuit-status") as response:
            if response.status_code == 200:
                try:
                    status = response.json()
                    logger.info(f"Circuit Status - Inventory: {status['inventory_circuit']['state']}, "
                              f"Payment: {status['payment_circuit']['state']}")
                    response.success()
                except Exception as e:
                    response.failure(f"Failed to parse circuit status: {e}")


class NormalOperationUser(ResilienceTestUser):
    """
    Scenario 1: Normal Operation
    Tests the system under normal conditions without any chaos
    """
    wait_time = between(0.5, 2)
    
    @task(10)
    def create_order(self):
        """Create valid orders"""
        self.create_valid_order()
    
    @task(2)
    def get_order(self):
        """Retrieve order details"""
        self.get_order_details()
    
    @task(1)
    def check_circuits(self):
        """Monitor circuit breaker status"""
        self.check_circuit_status()


class FailFastUser(ResilienceTestUser):
    """
    Scenario 2: Fail Fast Demo
    Tests input validation (should fail fast without calling downstream services)
    """
    wait_time = between(0.5, 1.5)
    
    @task(3)
    def empty_order(self):
        """Test empty order - should fail fast"""
        payload = {"items": []}
        with self.client.post("/order/create", json=payload, catch_response=True, 
                            name="/order/create [empty]") as response:
            if response.status_code == 400:
                response.success()
            else:
                response.failure(f"Expected 400, got {response.status_code}")
    
    @task(3)
    def invalid_quantity(self):
        """Test invalid quantity - should fail fast"""
        payload = {
            "items": [
                {"item_id": "item-1", "quantity": 0, "price": 999.99}
            ]
        }
        with self.client.post("/order/create", json=payload, catch_response=True,
                            name="/order/create [invalid quantity]") as response:
            if response.status_code == 400:
                response.success()
            else:
                response.failure(f"Expected 400, got {response.status_code}")
    
    @task(3)
    def invalid_price(self):
        """Test invalid price - should fail fast"""
        payload = {
            "items": [
                {"item_id": "item-1", "quantity": 1, "price": -10}
            ]
        }
        with self.client.post("/order/create", json=payload, catch_response=True,
                            name="/order/create [invalid price]") as response:
            if response.status_code == 400:
                response.success()
            else:
                response.failure(f"Expected 400, got {response.status_code}")
    
    @task(1)
    def missing_item_id(self):
        """Test missing item_id - should fail fast"""
        payload = {
            "items": [
                {"item_id": "", "quantity": 1, "price": 99.99}
            ]
        }
        with self.client.post("/order/create", json=payload, catch_response=True,
                            name="/order/create [missing item_id]") as response:
            if response.status_code == 400:
                response.success()
            else:
                response.failure(f"Expected 400, got {response.status_code}")


class CircuitBreakerInventoryUser(ResilienceTestUser):
    """
    Scenario 3: Circuit Breaker Demo (Inventory Failures)
    Tests circuit breaker with inventory service failures
    """
    wait_time = between(1, 3)
    
    @task(8)
    def create_order_with_failures(self):
        """Create orders that may fail due to inventory chaos"""
        self.create_valid_order()
    
    @task(2)
    def check_circuits(self):
        """Frequently check circuit status to observe state transitions"""
        self.check_circuit_status()


class CircuitBreakerPaymentUser(ResilienceTestUser):
    """
    Scenario 4: Circuit Breaker Demo (Payment Failures)
    Tests circuit breaker with payment service failures and slow responses
    """
    wait_time = between(1, 3)
    
    @task(8)
    def create_order_with_failures(self):
        """Create orders that may fail due to payment chaos"""
        self.create_valid_order()
    
    @task(2)
    def check_circuits(self):
        """Frequently check circuit status to observe state transitions"""
        self.check_circuit_status()


class BulkheadUser(ResilienceTestUser):
    """
    Scenario 5: Bulkhead Demo
    Tests bulkhead pattern by sending many concurrent requests
    """
    wait_time = constant(0)  # No wait time for maximum concurrency
    
    @task(10)
    def create_order_concurrent(self):
        """Create orders to test bulkhead limits (max 10 concurrent)"""
        self.create_valid_order()


class CombinedChaosUser(ResilienceTestUser):
    """
    Scenario 6: Combined Chaos
    Tests all resilience patterns with multiple failure modes enabled
    """
    wait_time = between(0.5, 2)
    
    @task(10)
    def create_order_with_all_chaos(self):
        """Create orders with all chaos modes enabled"""
        self.create_valid_order()
    
    @task(2)
    def get_order(self):
        """Retrieve order details"""
        self.get_order_details()
    
    @task(1)
    def check_circuits(self):
        """Monitor circuit breaker status"""
        self.check_circuit_status()


class MixedWorkloadUser(ResilienceTestUser):
    """
    Mixed workload - combines valid and invalid requests
    Useful for realistic load testing
    """
    wait_time = between(0.5, 2)
    
    @task(10)
    def create_valid_order(self):
        """Create valid orders"""
        self.create_valid_order()
    
    @task(2)
    def create_invalid_order(self):
        """Create invalid orders (fail fast)"""
        invalid_payloads = [
            {"items": []},
            {"items": [{"item_id": "item-1", "quantity": 0, "price": 999.99}]},
            {"items": [{"item_id": "item-1", "quantity": 1, "price": -10}]},
        ]
        payload = random.choice(invalid_payloads)
        self.client.post("/order/create", json=payload, name="/order/create [invalid]")
    
    @task(3)
    def get_order(self):
        """Retrieve order details"""
        self.get_order_details()
    
    @task(1)
    def check_circuits(self):
        """Monitor circuit breaker status"""
        self.check_circuit_status()


# Custom Load Test Shapes

class StepLoadShape(LoadTestShape):
    """
    Step load pattern: gradually increase load to test system behavior
    
    Stage 1 (0-60s):   10 users
    Stage 2 (60-120s): 25 users
    Stage 3 (120-180s): 50 users
    Stage 4 (180-240s): 75 users
    Stage 5 (240-300s): 100 users
    """
    
    stages = [
        {"duration": 60, "users": 10, "spawn_rate": 2},
        {"duration": 120, "users": 25, "spawn_rate": 3},
        {"duration": 180, "users": 50, "spawn_rate": 5},
        {"duration": 240, "users": 75, "spawn_rate": 5},
        {"duration": 300, "users": 100, "spawn_rate": 5},
    ]
    
    def tick(self):
        run_time = self.get_run_time()
        
        for stage in self.stages:
            if run_time < stage["duration"]:
                return (stage["users"], stage["spawn_rate"])
        
        return None  # Stop test after all stages


class SpikeLoadShape(LoadTestShape):
    """
    Spike load pattern: sudden increase to test bulkhead and circuit breaker
    
    Stage 1 (0-30s):   10 users (baseline)
    Stage 2 (30-60s):  100 users (spike)
    Stage 3 (60-90s):  10 users (recovery)
    Stage 4 (90-120s): 100 users (second spike)
    Stage 5 (120-150s): 10 users (recovery)
    """
    
    stages = [
        {"duration": 30, "users": 10, "spawn_rate": 2},
        {"duration": 60, "users": 100, "spawn_rate": 20},
        {"duration": 90, "users": 10, "spawn_rate": 10},
        {"duration": 120, "users": 100, "spawn_rate": 20},
        {"duration": 150, "users": 10, "spawn_rate": 10},
    ]
    
    def tick(self):
        run_time = self.get_run_time()
        
        for stage in self.stages:
            if run_time < stage["duration"]:
                return (stage["users"], stage["spawn_rate"])
        
        return None


# Event hooks for chaos control and monitoring

@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """Initialize test environment"""
    logger.info("=" * 60)
    logger.info("🎭 Resilience Patterns Load Test Starting")
    logger.info("=" * 60)
    logger.info(f"Target: {environment.host}")
    logger.info("")
    logger.info("Available User Classes:")
    logger.info("  - NormalOperationUser: Normal operation (no chaos)")
    logger.info("  - FailFastUser: Input validation tests")
    logger.info("  - CircuitBreakerInventoryUser: Inventory circuit breaker")
    logger.info("  - CircuitBreakerPaymentUser: Payment circuit breaker")
    logger.info("  - BulkheadUser: Bulkhead pattern (high concurrency)")
    logger.info("  - CombinedChaosUser: All chaos modes")
    logger.info("  - MixedWorkloadUser: Realistic mixed workload")
    logger.info("")
    logger.info("💡 Tip: Use --user-classes to run specific user classes")
    logger.info("=" * 60)


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """Cleanup after test"""
    logger.info("=" * 60)
    logger.info("✅ Resilience Patterns Load Test Complete")
    logger.info("=" * 60)
    logger.info("")
    logger.info("📊 View metrics at:")
    logger.info("   - Prometheus: http://localhost:9090")
    logger.info("   - Grafana:    http://localhost:3000")
    logger.info("   - Locust UI:  http://localhost:8089")
    logger.info("")
    logger.info("💡 Remember to disable chaos if enabled:")
    logger.info("   curl -X POST http://localhost:8081/chaos/inventory/disable")
    logger.info("   curl -X POST http://localhost:8082/chaos/payment/disable")
    logger.info("=" * 60)


# Helper class for direct chaos control from Locust

class ChaosController:
    """Helper class to control chaos engineering endpoints"""
    
    INVENTORY_SERVICE = "http://localhost:8081"
    PAYMENT_SERVICE = "http://localhost:8082"
    
    @staticmethod
    def enable_inventory_chaos():
        """Enable inventory service chaos (30% failure rate)"""
        import requests
        try:
            resp = requests.post(f"{ChaosController.INVENTORY_SERVICE}/chaos/inventory/enable", timeout=5)
            logger.info(f"✅ Inventory chaos enabled: {resp.json()}")
        except Exception as e:
            logger.error(f"❌ Failed to enable inventory chaos: {e}")
    
    @staticmethod
    def disable_inventory_chaos():
        """Disable inventory service chaos"""
        import requests
        try:
            resp = requests.post(f"{ChaosController.INVENTORY_SERVICE}/chaos/inventory/disable", timeout=5)
            logger.info(f"✅ Inventory chaos disabled: {resp.json()}")
        except Exception as e:
            logger.error(f"❌ Failed to disable inventory chaos: {e}")
    
    @staticmethod
    def enable_inventory_slow():
        """Enable inventory slow mode (2-5 second delays)"""
        import requests
        try:
            resp = requests.post(f"{ChaosController.INVENTORY_SERVICE}/chaos/inventory/slow", timeout=5)
            logger.info(f"✅ Inventory slow mode enabled: {resp.json()}")
        except Exception as e:
            logger.error(f"❌ Failed to enable inventory slow mode: {e}")
    
    @staticmethod
    def enable_payment_chaos():
        """Enable payment service chaos (40% failure rate)"""
        import requests
        try:
            resp = requests.post(f"{ChaosController.PAYMENT_SERVICE}/chaos/payment/enable", timeout=5)
            logger.info(f"✅ Payment chaos enabled: {resp.json()}")
        except Exception as e:
            logger.error(f"❌ Failed to enable payment chaos: {e}")
    
    @staticmethod
    def disable_payment_chaos():
        """Disable payment service chaos"""
        import requests
        try:
            resp = requests.post(f"{ChaosController.PAYMENT_SERVICE}/chaos/payment/disable", timeout=5)
            logger.info(f"✅ Payment chaos disabled: {resp.json()}")
        except Exception as e:
            logger.error(f"❌ Failed to disable payment chaos: {e}")
    
    @staticmethod
    def enable_payment_slow():
        """Enable payment slow mode (5-10 second delays)"""
        import requests
        try:
            resp = requests.post(f"{ChaosController.PAYMENT_SERVICE}/chaos/payment/slow", timeout=5)
            logger.info(f"✅ Payment slow mode enabled: {resp.json()}")
        except Exception as e:
            logger.error(f"❌ Failed to enable payment slow mode: {e}")
    
    @staticmethod
    def disable_all_chaos():
        """Disable all chaos modes"""
        logger.info("🔄 Disabling all chaos modes...")
        ChaosController.disable_inventory_chaos()
        ChaosController.disable_payment_chaos()
        logger.info("✅ All chaos modes disabled")
    
    @staticmethod
    def enable_combined_chaos():
        """Enable all chaos modes for combined chaos scenario"""
        logger.info("🔄 Enabling combined chaos...")
        ChaosController.enable_inventory_chaos()
        ChaosController.enable_inventory_slow()
        ChaosController.enable_payment_chaos()
        ChaosController.enable_payment_slow()
        logger.info("✅ Combined chaos enabled")


if __name__ == "__main__":
    """
    Example usage scenarios:
    
    # Normal operation test
    locust -f locustfile.py --host http://localhost:8080 --users 50 --spawn-rate 5 -t 120s
    
    # Fail fast validation test
    locust -f locustfile.py --host http://localhost:8080 --user-classes FailFastUser --users 20 --spawn-rate 5 -t 60s
    
    # Circuit breaker test (enable inventory chaos first)
    # curl -X POST http://localhost:8081/chaos/inventory/enable
    locust -f locustfile.py --host http://localhost:8080 --user-classes CircuitBreakerInventoryUser --users 30 --spawn-rate 3 -t 180s
    
    # Bulkhead test (enable payment slow mode first)
    # curl -X POST http://localhost:8082/chaos/payment/slow
    locust -f locustfile.py --host http://localhost:8080 --user-classes BulkheadUser --users 20 --spawn-rate 20 -t 60s
    
    # Combined chaos test (enable all chaos modes first)
    # curl -X POST http://localhost:8081/chaos/inventory/enable
    # curl -X POST http://localhost:8081/chaos/inventory/slow
    # curl -X POST http://localhost:8082/chaos/payment/enable
    # curl -X POST http://localhost:8082/chaos/payment/slow
    locust -f locustfile.py --host http://localhost:8080 --user-classes CombinedChaosUser --users 40 --spawn-rate 5 -t 300s
    
    # Mixed workload with step load
    locust -f locustfile.py --host http://localhost:8080 --user-classes MixedWorkloadUser --shape StepLoadShape
    
    # Spike load test
    locust -f locustfile.py --host http://localhost:8080 --user-classes MixedWorkloadUser --shape SpikeLoadShape
    """
    print(__doc__)
