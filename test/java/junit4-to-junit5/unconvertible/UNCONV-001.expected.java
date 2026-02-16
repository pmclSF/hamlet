import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Rule;
import org.junit.rules.ExpectedException;

public class ExpectedExceptionRuleTest {
    // HAMLET-TODO [UNCONVERTIBLE-RULE]: JUnit 4 @Rule/@ClassRule has no direct JUnit 5 equivalent
    // Original: @Rule
    // Manual action required: Use `assertThrows()` instead
    @Rule
    public ExpectedException thrown = ExpectedException.none();

    @Test
    public void testException() {
        thrown.expect(IllegalArgumentException.class);
        thrown.expectMessage("invalid");
        doSomething();
    }
}
