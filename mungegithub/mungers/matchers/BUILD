package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_test(
    name = "go_default_test",
    srcs = [
        "command_test.go",
        "interactions_test.go",
        "notification_test.go",
        "pinger_test.go",
    ],
    data = [
        "//mungegithub:configs",
    ],
    library = ":go_default_library",
    deps = ["//vendor/github.com/google/go-github/github:go_default_library"],
)

go_library(
    name = "go_default_library",
    srcs = [
        "command.go",
        "interactions.go",
        "item.go",
        "items.go",
        "matchers.go",
        "notification.go",
        "operators.go",
        "pinger.go",
    ],
    deps = [
        "//mungegithub/github:go_default_library",
        "//vendor/github.com/google/go-github/github:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//mungegithub/mungers/matchers/comment:all-srcs",
        "//mungegithub/mungers/matchers/event:all-srcs",
    ],
    tags = ["automanaged"],
)
