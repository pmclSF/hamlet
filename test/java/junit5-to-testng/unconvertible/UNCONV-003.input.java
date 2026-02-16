import org.junit.jupiter.api.RepeatedTest;

public class RepeatedTestTest {
    @RepeatedTest(5)
    public void testRepeated() {
        assert true;
    }
}
