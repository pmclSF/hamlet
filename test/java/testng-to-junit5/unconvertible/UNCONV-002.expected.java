import org.junit.jupiter.api.Test;

public class PriorityTest {
    // HAMLET-TODO [UNCONVERTIBLE-PRIORITY]: TestNG priority has no direct JUnit 5 equivalent
    // Original: @Test(priority = 1)
    // Manual action required: Use @Order annotation with @TestMethodOrder(OrderAnnotation.class)
    @Test(priority = 1)
    public void testFirst() {
        assert true;
    }

    // HAMLET-TODO [UNCONVERTIBLE-PRIORITY]: TestNG priority has no direct JUnit 5 equivalent
    // Original: @Test(priority = 2)
    // Manual action required: Use @Order annotation with @TestMethodOrder(OrderAnnotation.class)
    @Test(priority = 2)
    public void testSecond() {
        assert true;
    }
}
