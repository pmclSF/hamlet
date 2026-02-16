import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class MultiAssertTest {
    @Test
    public void testMultiple() {
        Assertions.assertEquals(1, getOne());
        Assertions.assertTrue(isTrue());
        Assertions.assertNotNull(getObj());
        Assertions.assertFalse(isFalse());
    }
}
