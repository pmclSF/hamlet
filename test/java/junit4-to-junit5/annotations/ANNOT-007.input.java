import org.junit.Test;
import org.junit.experimental.categories.Category;

@Category(SlowTests.class)
public class CategoryTest {

    @Test
    public void testSlow() {
        assertTrue(true);
    }
}
