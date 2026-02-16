import org.testng.annotations.Factory;
import org.junit.jupiter.api.Test;

public class FactoryTest {
    // HAMLET-TODO [UNCONVERTIBLE-FACTORY]: TestNG @Factory has no direct JUnit 5 equivalent
    // Original: @Factory
    // Manual action required: Use @ParameterizedTest or @TestFactory in JUnit 5
    @Factory
    public Object[] createTests() {
        return new Object[] { new FactoryTest() };
    }

    @Test
    public void testSomething() {
        assert true;
    }
}
