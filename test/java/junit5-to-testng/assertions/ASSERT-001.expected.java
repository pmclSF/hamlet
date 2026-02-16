import org.testng.annotations.Test;
import org.testng.Assert;

public class AssertEqualsTest {
    @Test
    public void testEquals() {
        Assert.assertEquals(getAnswer(), 42);
    }
}
