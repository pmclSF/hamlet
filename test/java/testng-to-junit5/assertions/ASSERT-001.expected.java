import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class AssertEqualsTest {
    @Test
    public void testEquals() {
        Assertions.assertEquals(42, getResult());
    }
}
