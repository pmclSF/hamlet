import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;

public class IgnoredTest {

    @Disabled
    @Test
    public void testSkipped() {
        fail("This test should be skipped");
    }

    @Test
    public void testActive() {
        assertTrue(true);
    }
}
