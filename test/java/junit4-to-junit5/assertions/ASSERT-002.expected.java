import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class AssertBooleanTest {
    @Test
    public void testBoolean() {
        Assertions.assertTrue(isValid());
        Assertions.assertFalse(isEmpty());
    }
}
