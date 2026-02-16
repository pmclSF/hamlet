import org.junit.Test;
import org.junit.Assert;

public class AssertArrayTest {
    @Test
    public void testArray() {
        int[] expected = {1, 2, 3};
        int[] actual = getArray();
        Assert.assertArrayEquals(expected, actual);
    }
}
