import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertTimeout;
import java.time.Duration;

public class TimeoutTest {
    @Test
    public void testTimeout() {
        assertTimeout(Duration.ofMillis(1000), () -> {
            longRunningOperation();
        });
    }
}
