package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = [
        "bool.go",
        "count.go",
        "duration.go",
        "flag.go",
        "float32.go",
        "float64.go",
        "golangflag.go",
        "int.go",
        "int32.go",
        "int64.go",
        "int8.go",
        "int_slice.go",
        "ip.go",
        "ipmask.go",
        "ipnet.go",
        "string.go",
        "string_slice.go",
        "uint.go",
        "uint16.go",
        "uint32.go",
        "uint64.go",
        "uint8.go",
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
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
