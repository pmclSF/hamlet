import org.testng.annotations.Listeners;
import org.junit.jupiter.api.Test;

// HAMLET-TODO [UNCONVERTIBLE-LISTENERS]: TestNG @Listeners has no direct JUnit 5 equivalent
// Original: @Listeners(MyTestListener.class)
// Manual action required: Use @ExtendWith with JUnit 5 extension instead
@Listeners(MyTestListener.class)
public class ListenersTest {
    @Test
    public void testSomething() {
        assert true;
    }
}
