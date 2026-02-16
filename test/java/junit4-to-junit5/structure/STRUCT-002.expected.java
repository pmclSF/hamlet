import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class MultiMethodTest {
    @Test
    public void testFirst() {
        Assertions.assertEquals(1, 1);
    }

    @Test
    public void testSecond() {
        Assertions.assertEquals(2, 2);
    }

    @Test
    public void testThird() {
        Assertions.assertNotNull("hello");
    }
}
