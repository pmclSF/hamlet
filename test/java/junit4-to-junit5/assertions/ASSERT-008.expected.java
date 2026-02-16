import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class AssertTrueMessageTest {
    @Test
    public void testTrueMessage() {
        Assertions.assertTrue(isValid(), "should be valid");
    }
}
