import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertThrows;

public class ExpectedExceptionTest {
    @Test
    public void testThrows() {
        assertThrows(IllegalArgumentException.class, () -> {
            Integer.parseInt("not a number");
        });
    }
}
