# Inertia.js Go (Echo) Adapter

The Inertia.js server-side adapter for Go. Visit [inertiajs.com](https://inertiajs.com) to learn more.

---

This module has fixed some bugs. Using `Vite` instead of `Laravel-Mix`
> Original Source Code from here: https://github.com/elipZis/inertia-echo .<br />
> Stripped off everything except `./LICENSE`.

## Usage
Create a new Echo instance and register the Inertia middleware with it

```golang
func main() {
    e := echo.New()
    e.Use(inertia.Middleware(e))
    e.Static("/", "./public")

    e.GET("/", hello).Name = "hello"

    e.Logger.Fatal(e.Start("127.0.1.1:3000"))
}

func hello(c echo.Context) error {
    return c.Render(http.StatusOK, "Welcome", map[string]interface{}{})
}
```

The internal template renderer of Inertia-Echo checks whether a fresh full base-site has to be returned or only the reduced Inertia response.

> For more examples refer to the `demo` branch at https://github.com/theartefak/demo-inertia-echo

## License

- **elipZis** (https://github.com/elipZis/inertia-echo/blob/master/README.md).
- **Echo** (https://github.com/labstack/echo/blob/master/LICENSE).
and many more.

This open-sourced software is licensed under the [MIT license](LICENSE).
