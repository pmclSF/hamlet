import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;

public class DisabledTest {
    @Disabled
    @Test
    public void testSkipped() {
        assert false;
    }
}
