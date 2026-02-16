// TestNG uses @Test(groups = {...}) instead of @Tag
import org.testng.annotations.Test;

public class TagTest {
    @Test(groups = {"slow"})
    public void testSlow() {
        assert true;
    }
}
