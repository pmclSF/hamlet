import org.springframework.web.bind.annotation.*;

@RestController
public class AuthController {

    @PostMapping("/api/login")
    public String login() {
        return "token";
    }

    @GetMapping("/api/profile")
    public String profile() {
        return "profile";
    }

    public void internalMethod() {
        // not a route
    }
}
