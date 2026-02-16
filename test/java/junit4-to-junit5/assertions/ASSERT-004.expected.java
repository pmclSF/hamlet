import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class AssertSameTest {
    @Test
    public void testSame() {
        Object obj = new Object();
        Assertions.assertSame(obj, obj);
    }
}
