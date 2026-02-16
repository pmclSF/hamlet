import org.junit.jupiter.api.*;

public class DisabledTest {
    @Disabled
    @Test
    public void testSkipped() {
        assert false;
    }
}
