import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTimeout;
import java.time.Duration;

public class CombinedTest {
    @Test
    public void testExpected() {
        assertThrows(RuntimeException.class, () -> {
            throwRuntime();
        });
    }

    @Test
    public void testTimed() {
        assertTimeout(Duration.ofMillis(500), () -> {
            quickOperation();
        });
    }
}
