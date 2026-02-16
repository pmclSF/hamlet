import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Rule;
import org.junit.rules.TemporaryFolder;

public class TempFolderRuleTest {
    // HAMLET-TODO [UNCONVERTIBLE-RULE]: JUnit 4 @Rule/@ClassRule has no direct JUnit 5 equivalent
    // Original: @Rule
    // Manual action required: Use `@TempDir` annotation
    @Rule
    public TemporaryFolder folder = new TemporaryFolder();

    @Test
    public void testTempFile() throws Exception {
        folder.newFile("test.txt");
    }
}
