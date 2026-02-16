import random

from locust import FastHttpUser, task, between


class OrderServiceUser(FastHttpUser):
    """Simulates user traffic for order creation and retrieval."""
    wait_time = between(0.25, 1)
    # Store created order IDs for retrieval
    order_ids = []
    
    items = [
        {"item_id": "item-1", "quantity": 1, "price": 999.99},
        {"item_id": "item-2", "quantity": 2, "price": 29.99},
        {"item_id": "item-3", "quantity": 1, "price": 79.99},
        {"item_id": "item-4", "quantity": 1, "price": 299.99},
        {"item_id": "item-5", "quantity": 3, "price": 149.99}
    ]
    
    @task(3)
    def create_order(self):
        # Randomly select 1-3 items for the order
        num_items = random.randint(1, 3)
        selected_items = random.sample(self.items, num_items)
        
        payload = {"items": selected_items}
        
        with self.client.post("/order/create", json=payload, catch_response=True, name="/order/create") as response:
            if response.status_code == 200:
                try:
                    order_data = response.json()
                    order_id = order_data.get("order_id")
                    if order_id:
                        # Store order ID for later retrieval
                        if len(self.order_ids) < 100:  # Limit stored IDs
                            self.order_ids.append(order_id)
                    response.success()
                except:
                    response.failure("Failed to parse response")
            else:
                response.failure(f"Got status {response.status_code}")
        
    @task(1)
    def get_order(self):
        if self.order_ids:
            # Get a random order ID from stored ones
            order_id = random.choice(self.order_ids)
            self.client.get(f"/order/{order_id}", name="/order/:orderId")
        else:
            # If no orders created yet, try a dummy ID (will likely 404, but that's ok for testing)
            self.client.get("/order/dummy-order-id", name="/order/:orderId")