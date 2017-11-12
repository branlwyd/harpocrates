load("@io_bazel_rules_go//go:def.bzl", "go_prefix", "go_binary", "go_library", "go_test")

go_prefix("github.com/BranLwyd/harpocrates")

##
## Binaries
##
go_binary(
    name = "harpd",
    srcs = ["harpd.go"],
    deps = [
        "//:counter",
        "//:server",
        "//handler",
        "@org_golang_x_crypto//acme/autocert:go_default_library",
    ],
)

go_binary(
    name = "harpd_debug",
    srcs = ["harpd_debug.go"],
    deps = [
        "//:counter",
        "//:debug_assets",
        "//:server",
        "//handler",
    ],
)

##
## Libraries
##
go_library(
    name = "alert",
    srcs = ["alert.go"],
)

go_library(
    name = "counter",
    srcs = ["counter.go"],
)

go_library(
    name = "password",
    srcs = ["password.go"],
    visibility = ["//handler:__pkg__"],
    deps = [
        "@org_golang_x_crypto//openpgp:go_default_library",
        "@org_golang_x_crypto//ripemd160:go_default_library",
    ],
)

go_test(
    name = "password_test",
    timeout = "short",
    srcs = ["password_test.go"],
    library = ":password",
    deps = [
        "@org_golang_x_crypto//openpgp:go_default_library",
    ],
)

go_library(
    name = "rate",
    srcs = ["rate.go"],
    visibility = ["//handler:__pkg__"],
)

go_library(
    name = "server",
    srcs = ["server.go"],
    deps = [
        "//:alert",
        "//:counter",
        "//:session",
        "//handler",
    ],
)

go_library(
    name = "session",
    srcs = ["session.go"],
    visibility = ["//handler:__pkg__"],
    deps = [
        "//:alert",
        "//:counter",
        "//:password",
        "//:rate",
        "@com_github_tstranex_u2f//:go_default_library",
        "@org_golang_x_crypto//openpgp:go_default_library",
        "@org_golang_x_crypto//openpgp/packet:go_default_library",
    ],
)

##
## Static assets
##
filegroup(
    name = "assets_files",
    srcs = glob(
        ["assets/**/*"],
        exclude = ["assets/debug/**/*"],
    ),
)

genrule(
    name = "assets_go",
    srcs = [":assets_files"],
    outs = ["assets.go"],
    cmd = "go-bindata -o $@ --nomemcopy --nocompress --pkg=assets --prefix=assets/ $(locations :assets_files)",
)

go_library(
    name = "assets",
    srcs = ["assets.go"],
    visibility = ["//handler:__pkg__"],
)

filegroup(
    name = "debug_assets_files",
    srcs = glob(["assets/debug/**/*"]),
)

genrule(
    name = "debug_assets_go",
    srcs = [":debug_assets_files"],
    outs = ["debug_assets.go"],
    cmd = "go-bindata -o $@ --nomemcopy --nocompress --pkg=debug_assets --prefix=assets/ $(locations :debug_assets_files)",
)

go_library(
    name = "debug_assets",
    srcs = ["debug_assets.go"],
)