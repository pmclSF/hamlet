import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

public class NestedTest {
    @Nested
    class InnerTest {
        @Test
        public void testInner() {
            assert true;
        }
    }
}
