import org.junit.jupiter.api.*;
import org.junit.jupiter.api.Assertions;

public class CombinedTest {
    @BeforeEach
    public void setUp() {
        System.out.println("setup");
    }

    @AfterEach
    public void tearDown() {
        System.out.println("teardown");
    }

    @Test
    public void testBasic() {
        Assertions.assertTrue(true);
    }

    @Disabled
    @Test
    public void testSkipped() {
        assert false;
    }
}
