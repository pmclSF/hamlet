import org.junit.Test;
import org.junit.Rule;
import org.junit.rules.ErrorCollector;

public class ErrorCollectorRuleTest {
    @Rule
    public ErrorCollector collector = new ErrorCollector();

    @Test
    public void testMultipleErrors() {
        collector.checkThat(1, is(1));
        collector.checkThat(2, is(3));
    }
}
