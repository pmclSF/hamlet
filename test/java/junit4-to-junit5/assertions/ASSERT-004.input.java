import org.junit.Test;
import org.junit.Assert;

public class AssertSameTest {
    @Test
    public void testSame() {
        Object obj = new Object();
        Assert.assertSame(obj, obj);
    }
}
