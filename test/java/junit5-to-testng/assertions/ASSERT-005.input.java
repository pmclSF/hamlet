import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class MultiAssertTest {
    @Test
    public void testMultiple() {
        Assertions.assertNotNull(getResult());
        Assertions.assertEquals(42, getResult().getValue());
        Assertions.assertTrue(getResult().isValid());
    }
}
