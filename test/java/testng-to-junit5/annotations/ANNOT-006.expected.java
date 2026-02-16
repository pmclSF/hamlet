import org.junit.jupiter.api.*;

public class GroupTest {
    @Tag("slow")
    @Test
    public void testSlow() {
        assert true;
    }
}
