import org.testng.annotations.*;

public class DisabledTest {
    @Test(enabled = false)
    public void testSkipped() {
        assert false;
    }
}
