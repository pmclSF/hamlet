import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class OuterTest {
    @Test
    public void testOuter() {
        Assertions.assertTrue(true);
    }

    public static class InnerTest {
        @Test
        public void testInner() {
            Assertions.assertEquals(1, 1);
        }
    }
}
