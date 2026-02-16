import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Rule;
import org.junit.rules.TestName;

public class TestNameRuleTest {
    // HAMLET-TODO [UNCONVERTIBLE-RULE]: JUnit 4 @Rule/@ClassRule has no direct JUnit 5 equivalent
    // Original: @Rule
    // Manual action required: Use `TestInfo` parameter injection
    @Rule
    public TestName testName = new TestName();

    @Test
    public void testGetName() {
        System.out.println(testName.getMethodName());
    }
}
