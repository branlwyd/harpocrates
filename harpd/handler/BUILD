load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "handler",
    srcs = [
        "auth.go",
        "content.go",
        "logout.go",
        "mfa.go",
        "misc.go",
        "password.go",
        "search.go",
    ],
    importpath = "github.com/BranLwyd/harpocrates/harpd/handler",
    visibility = ["//harpd:__pkg__"],
    deps = [
        "//harpd:assets",
        "//harpd:rate",
        "//harpd:session",
        "//secret",
        "@cc_mvdan_xurls//:go_default_library",
        "@com_github_e3b0c442_warp//:go_default_library",
        "@org_golang_x_text//collate:go_default_library",
        "@org_golang_x_text//language:go_default_library",
        "@org_golang_x_text//search:go_default_library",
    ],
)
