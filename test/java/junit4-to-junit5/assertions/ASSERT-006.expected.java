import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class AssertMessageTest {
    @Test
    public void testMessage() {
        Assertions.assertEquals(42, getResult(), "values should match");
    }
}
