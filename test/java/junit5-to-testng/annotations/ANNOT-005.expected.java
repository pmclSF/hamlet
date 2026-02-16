// TestNG uses @Test(enabled = false) instead of @Disabled
import org.testng.annotations.Test;

public class DisabledTest {
    @Test(enabled = false)
    public void testSkipped() {
        assert false;
    }
}
