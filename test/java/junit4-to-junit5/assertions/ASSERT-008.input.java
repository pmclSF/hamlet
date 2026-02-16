import org.junit.Test;
import org.junit.Assert;

public class AssertTrueMessageTest {
    @Test
    public void testTrueMessage() {
        Assert.assertTrue("should be valid", isValid());
    }
}
