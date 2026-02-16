import org.junit.jupiter.api.TestFactory;

public class TestFactoryTest {
    // HAMLET-TODO [UNCONVERTIBLE-TEST-FACTORY]: JUnit 5 @TestFactory has no TestNG equivalent
    // Original: @TestFactory
    // Manual action required: Use @DataProvider or @Factory in TestNG
    @TestFactory
    public void testFactory() {
        assert true;
    }
}
