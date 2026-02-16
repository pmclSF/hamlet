import org.junit.Test;
import org.junit.Assume;

public class AssumeTest {
    @Test
    public void testAssume() {
        Assume.assumeTrue(isLinux());
        assert true;
    }
}
