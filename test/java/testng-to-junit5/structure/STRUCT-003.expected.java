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
    public void testAdd() {
        Assertions.assertEquals(1 + 1, 2);
        Assertions.assertTrue(true);
    }

    @Test
    public void testSubtract() {
        Assertions.assertEquals(1 - 1, 0);
    }
}
