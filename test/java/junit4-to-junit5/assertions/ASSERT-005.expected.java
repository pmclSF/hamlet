import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class AssertArrayTest {
    @Test
    public void testArray() {
        int[] expected = {1, 2, 3};
        int[] actual = getArray();
        Assertions.assertArrayEquals(expected, actual);
    }
}
