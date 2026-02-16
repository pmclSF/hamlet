import org.junit.Ignore;
import org.junit.Test;

public class IgnoredTest {

    @Ignore
    @Test
    public void testSkipped() {
        fail("This test should be skipped");
    }

    @Test
    public void testActive() {
        assertTrue(true);
    }
}
