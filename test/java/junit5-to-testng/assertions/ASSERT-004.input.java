import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class AssertNullTest {
    @Test
    public void testNull() {
        Assertions.assertNull(getResult());
    }
}
