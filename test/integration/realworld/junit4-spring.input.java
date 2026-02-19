// JUnit 4 test for a Spring service layer
// Inspired by real-world Spring Boot service tests with dependency injection

package com.example.shop.service;

import org.junit.Before;
import org.junit.After;
import org.junit.Test;
import org.junit.Assert;
import org.junit.runner.RunWith;
import org.springframework.test.context.junit4.SpringRunner;

import java.math.BigDecimal;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

@RunWith(SpringRunner.class)
public class OrderServiceTest {

    private OrderService orderService;
    private InMemoryOrderRepository orderRepository;
    private InMemoryProductRepository productRepository;

    @Before
    public void setUp() {
        orderRepository = new InMemoryOrderRepository();
        productRepository = new InMemoryProductRepository();
        productRepository.save(new Product(1L, "Widget", new BigDecimal("29.99")));
        productRepository.save(new Product(2L, "Gadget", new BigDecimal("49.99")));
        orderService = new OrderService(orderRepository, productRepository);
    }

    @After
    public void tearDown() {
        orderRepository.clear();
        productRepository.clear();
    }

    @Test
    public void createOrder_withValidItems_returnsOrderWithTotal() {
        List<OrderItem> items = Arrays.asList(
            new OrderItem(1L, 2),
            new OrderItem(2L, 1)
        );

        Order order = orderService.createOrder("customer-1", items);

        Assert.assertNotNull(order.getId());
        Assert.assertEquals("customer-1", order.getCustomerId());
        Assert.assertEquals(new BigDecimal("109.97"), order.getTotal());
        Assert.assertEquals(2, order.getItems().size());
    }

    @Test
    public void createOrder_withSingleItem_calculatesCorrectTotal() {
        List<OrderItem> items = Arrays.asList(new OrderItem(1L, 3));

        Order order = orderService.createOrder("customer-2", items);

        Assert.assertEquals(new BigDecimal("89.97"), order.getTotal());
    }

    @Test(expected = IllegalArgumentException.class)
    public void createOrder_withEmptyItems_throwsException() {
        orderService.createOrder("customer-3", Arrays.asList());
    }

    @Test(expected = ProductNotFoundException.class)
    public void createOrder_withInvalidProductId_throwsProductNotFound() {
        List<OrderItem> items = Arrays.asList(new OrderItem(999L, 1));
        orderService.createOrder("customer-4", items);
    }

    @Test
    public void findOrderById_withExistingOrder_returnsOrder() {
        List<OrderItem> items = Arrays.asList(new OrderItem(1L, 1));
        Order created = orderService.createOrder("customer-5", items);

        Optional<Order> found = orderService.findById(created.getId());

        Assert.assertTrue(found.isPresent());
        Assert.assertEquals(created.getId(), found.get().getId());
    }

    @Test
    public void findOrderById_withNonExistentId_returnsEmpty() {
        Optional<Order> found = orderService.findById("nonexistent-id");

        Assert.assertFalse(found.isPresent());
    }

    @Test(timeout = 1000)
    public void findOrdersByCustomer_performsWithinTimeLimit() {
        for (int i = 0; i < 100; i++) {
            List<OrderItem> items = Arrays.asList(new OrderItem(1L, 1));
            orderService.createOrder("bulk-customer", items);
        }

        List<Order> orders = orderService.findByCustomerId("bulk-customer");

        Assert.assertEquals(100, orders.size());
    }

    @Test
    public void cancelOrder_setsStatusToCancelled() {
        List<OrderItem> items = Arrays.asList(new OrderItem(2L, 1));
        Order order = orderService.createOrder("customer-6", items);

        orderService.cancelOrder(order.getId());

        Optional<Order> found = orderService.findById(order.getId());
        Assert.assertTrue(found.isPresent());
        Assert.assertEquals(OrderStatus.CANCELLED, found.get().getStatus());
    }
}
