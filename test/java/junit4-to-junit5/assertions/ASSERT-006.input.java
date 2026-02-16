import org.junit.Test;
import org.junit.Assert;

public class AssertMessageTest {
    @Test
    public void testMessage() {
        Assert.assertEquals("values should match", 42, getResult());
    }
}
