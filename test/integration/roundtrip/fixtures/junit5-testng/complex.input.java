package com.example.service;

import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

import java.math.BigDecimal;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertAll;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

@DisplayName("OrderService integration tests")
@Tag("integration")
class OrderServiceTest {

    private static DatabaseConnection db;
    private OrderService orderService;
    private InventoryService inventoryService;

    @BeforeAll
    static void initDatabase() {
        db = DatabaseConnection.createTestInstance();
        db.migrate();
    }

    @AfterAll
    static void tearDownDatabase() {
        db.close();
    }

    @BeforeEach
    void setUp() {
        db.truncateAll();
        inventoryService = new InventoryService(db);
        orderService = new OrderService(db, inventoryService);
        inventoryService.addStock("SKU-100", 50);
        inventoryService.addStock("SKU-200", 10);
    }

    @Test
    @DisplayName("should create an order with a unique identifier")
    void shouldCreateOrder() {
        Order order = orderService.placeOrder("customer-1", List.of(
            new OrderItem("SKU-100", 2, new BigDecimal("19.99"))
        ));
        assertNotNull(order.getId());
        assertEquals("PENDING", order.getStatus());
    }

    @Test
    @DisplayName("should calculate the order total correctly")
    void shouldCalculateTotal() {
        Order order = orderService.placeOrder("customer-1", List.of(
            new OrderItem("SKU-100", 3, new BigDecimal("10.00")),
            new OrderItem("SKU-200", 1, new BigDecimal("25.50"))
        ));
        assertEquals(new BigDecimal("55.50"), order.getTotal());
    }

    @Test
    @DisplayName("should reject an order when stock is insufficient")
    void shouldRejectInsufficientStock() {
        assertThrows(InsufficientStockException.class, () ->
            orderService.placeOrder("customer-1", List.of(
                new OrderItem("SKU-200", 100, new BigDecimal("5.00"))
            ))
        );
    }

    @Nested
    @DisplayName("order fulfillment")
    class Fulfillment {

        private Order pendingOrder;

        @BeforeEach
        void createPendingOrder() {
            pendingOrder = orderService.placeOrder("customer-1", List.of(
                new OrderItem("SKU-100", 1, new BigDecimal("19.99"))
            ));
        }

        @Test
        @DisplayName("should mark the order as shipped")
        void shouldMarkAsShipped() {
            Order shipped = orderService.ship(pendingOrder.getId(), "TRACK-001");
            assertAll(
                () -> assertEquals("SHIPPED", shipped.getStatus()),
                () -> assertEquals("TRACK-001", shipped.getTrackingNumber()),
                () -> assertNotNull(shipped.getShippedAt())
            );
        }

        @Test
        @DisplayName("should reduce inventory after shipping")
        void shouldReduceInventory() {
            orderService.ship(pendingOrder.getId(), "TRACK-002");
            int remaining = inventoryService.getStock("SKU-100");
            assertEquals(49, remaining);
        }

        @Test
        @DisplayName("should prevent shipping a cancelled order")
        @Tag("negative")
        void shouldNotShipCancelledOrder() {
            orderService.cancel(pendingOrder.getId());
            assertThrows(InvalidOrderStateException.class, () ->
                orderService.ship(pendingOrder.getId(), "TRACK-003")
            );
        }
    }

    @Nested
    @DisplayName("order cancellation")
    class Cancellation {

        @Test
        @DisplayName("should cancel a pending order")
        void shouldCancelPendingOrder() {
            Order order = orderService.placeOrder("customer-2", List.of(
                new OrderItem("SKU-100", 5, new BigDecimal("10.00"))
            ));
            Order cancelled = orderService.cancel(order.getId());
            assertEquals("CANCELLED", cancelled.getStatus());
        }

        @Test
        @DisplayName("should restore inventory after cancellation")
        void shouldRestoreInventory() {
            Order order = orderService.placeOrder("customer-2", List.of(
                new OrderItem("SKU-200", 3, new BigDecimal("5.00"))
            ));
            orderService.cancel(order.getId());
            assertEquals(10, inventoryService.getStock("SKU-200"));
        }
    }

    @ParameterizedTest
    @ValueSource(ints = {1, 5, 10, 25, 50})
    @DisplayName("should handle varying order quantities")
    void shouldHandleVaryingQuantities(int quantity) {
        Order order = orderService.placeOrder("customer-3", List.of(
            new OrderItem("SKU-100", quantity, new BigDecimal("10.00"))
        ));
        assertNotNull(order.getId());
        assertTrue(order.getTotal().compareTo(BigDecimal.ZERO) > 0);
    }
}
