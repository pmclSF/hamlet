import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class MultiMethodTest {
    @Test
    public void testFirst() {
        Assertions.assertTrue(true);
    }

    @Test
    public void testSecond() {
        Assertions.assertEquals(42, getResult());
    }

    @Test
    public void testThird() {
        Assertions.assertNotNull(getObj());
    }
}
