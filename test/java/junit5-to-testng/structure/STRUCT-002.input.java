import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class MultiMethodTest {
    @Test
    public void testFirst() {
        Assertions.assertEquals(1, 1);
    }

    @Test
    public void testSecond() {
        Assertions.assertTrue(true);
    }
}
