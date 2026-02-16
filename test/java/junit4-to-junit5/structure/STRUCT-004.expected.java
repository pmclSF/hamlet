import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class ChildTest extends BaseTest {
    @Test
    public void testChild() {
        Assertions.assertTrue(isReady());
    }
}
