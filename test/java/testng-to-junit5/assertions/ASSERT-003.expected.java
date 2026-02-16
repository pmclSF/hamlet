import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertThrows;

public class ExpectedExceptionsTest {
    @Test
    public void testException() {
        assertThrows(IllegalArgumentException.class, () -> {
            Integer.parseInt("not a number");
        });
    }
}
