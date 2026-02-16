import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assumptions;

public class AssumeTest {
    @Test
    public void testAssume() {
        Assumptions.assumeTrue(isLinux());
        assert true;
    }
}
