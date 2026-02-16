import org.junit.Test;
import org.junit.Assert;
import java.util.List;
import java.util.ArrayList;

public class MixedImportTest {
    @Test
    public void testList() {
        List<String> items = new ArrayList<>();
        items.add("hello");
        Assert.assertEquals(1, items.size());
    }
}
