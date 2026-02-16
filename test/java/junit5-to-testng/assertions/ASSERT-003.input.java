import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertThrows;

public class AssertThrowsTest {
    @Test
    public void testException() {
        assertThrows(IllegalArgumentException.class, () -> {
            throwIllegalArg();
        });
    }
}
