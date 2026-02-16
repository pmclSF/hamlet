import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Rule;
import org.junit.rules.ErrorCollector;

public class ErrorCollectorRuleTest {
    // HAMLET-TODO [UNCONVERTIBLE-RULE]: JUnit 4 @Rule/@ClassRule has no direct JUnit 5 equivalent
    // Original: @Rule
    // Manual action required: Use `assertAll()` for grouped assertions
    @Rule
    public ErrorCollector collector = new ErrorCollector();

    @Test
    public void testMultipleErrors() {
        collector.checkThat(1, is(1));
        collector.checkThat(2, is(3));
    }
}
