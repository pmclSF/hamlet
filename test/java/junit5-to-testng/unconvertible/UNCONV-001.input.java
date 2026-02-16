import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

public class DisplayNameTest {
    @DisplayName("A meaningful test name")
    @Test
    public void testSomething() {
        assert true;
    }
}
