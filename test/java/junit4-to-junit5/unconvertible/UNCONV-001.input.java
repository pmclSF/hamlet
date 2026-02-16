import org.junit.Test;
import org.junit.Rule;
import org.junit.rules.ExpectedException;

public class ExpectedExceptionRuleTest {
    @Rule
    public ExpectedException thrown = ExpectedException.none();

    @Test
    public void testException() {
        thrown.expect(IllegalArgumentException.class);
        thrown.expectMessage("invalid");
        doSomething();
    }
}
