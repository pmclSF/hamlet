import org.junit.Test;
import org.junit.Assert;

public class AssertNullTest {
    @Test
    public void testNull() {
        Assert.assertNull(getNull());
        Assert.assertNotNull(getObject());
    }
}
