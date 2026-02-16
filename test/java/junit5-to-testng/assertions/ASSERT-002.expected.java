import org.testng.annotations.Test;
import org.testng.Assert;

public class AssertTrueTest {
    @Test
    public void testTrue() {
        Assert.assertTrue(isValid());
    }
}
