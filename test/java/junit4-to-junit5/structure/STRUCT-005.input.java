import org.junit.Test;
import org.junit.Assert;

public class OuterTest {
    @Test
    public void testOuter() {
        Assert.assertTrue(true);
    }

    public static class InnerTest {
        @Test
        public void testInner() {
            Assert.assertEquals(1, 1);
        }
    }
}
