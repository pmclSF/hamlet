import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

public class TagTest {
    @Tag("slow")
    @Test
    public void testSlow() {
        assert true;
    }
}
