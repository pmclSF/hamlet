import org.junit.Test;
import org.junit.Before;
import org.junit.After;
import org.junit.Assert;
import org.junit.Ignore;
import java.util.List;
import java.util.ArrayList;

public class FullConversionTest {
    private List<String> items;

    @Before
    public void setUp() {
        items = new ArrayList<>();
    }

    @After
    public void tearDown() {
        items.clear();
    }

    @Test
    public void testAdd() {
        items.add("hello");
        Assert.assertEquals(1, items.size());
        Assert.assertTrue(items.contains("hello"));
    }

    @Ignore
    @Test
    public void testNotReady() {
        Assert.fail("not implemented");
    }
}
