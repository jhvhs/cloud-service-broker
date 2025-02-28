# Brokerpak introduction

This document will explain brokerpaks, what they are, the problems they solve, how to use them, and best practices around developing, maintaining, and using paks.

## What is a brokerpak?

To understand the how and why of brokerpaks, it's first important to understand the kind of problems that service brokers solve.

### A quick aside about service brokers

A _service broker_ provides an interface between a _service provider_ (e.g. GCP, Azure or AWS), and an _application platform_ (e.g. Kubernetes or Cloud Foundry).

There are two agents when it comes to talking about a service broker:
* Platform operators
    * The platform operators are responsible for configuring and managing the broker to meet the needs of their business, platform, and developers.
* Developers
    * The developers use the broker to provision and bind new services to their applications.

Therefore, a service broker is responsible for federating access between an _application provider_ and a _developer_ with respecting the wishes of the _platform_ and its _operators_.
Each of these parties influences the broker, its services, and its structure.

* Developers want lots of services that require minimal configuration yet give them enough control that they have independence.
* Operators need to make sure the services they expose are secure, follow regulatory constraints, can be billed correctly, are well supported, and won't be abused.
* Service providers are interested in providing lots of stable, generic services at a rapid pace.
* Service brokers serve the needs of the operators, developers, and platforms. They map the services out to match a variety of different business models, threat models, regulatory constraints, and use-cases.

Together, this means a service broker must:

* Provide many services for developers.
* Provide services granular enough that operators can control costs.
* Provide services robust enough that operators can control security.
* Provide services structured enough that operators can trust they'll be in compliance.
* Provide services configurable enough that developers will be happy.
* Map a single platform service into its N use-cases so operators can grant developers fine-grained access.
* Write documentation for each of those N use-cases.
* Be backwards compatible with all changes so developers and operators get seamless upgrades.
* Do all of the above at the rapid pace the platform grows at.

Some of these tasks can be automated, but many require deeper understanding of the platform and specific services to do correctly.

### How brokerpak solves these problems

The brokerpak is a (zip) package that contains bundled versions of [Terraform](https://terraform.io/intro/index.html), service definitions (as Terraform modules), Terraform providers, and (optionally) source code.
Brokerpaks are written as code, can be built as part of a CI/CD platform, and can be stored as build artifacts.
Operators can build and deploy their own brokerpaks along with those provided by the platform.
Operators can also modify the brokerpaks provided by the platform if they need to tailor the experience for their users.

Together this means that operators and developers can start to collaborate on developing new services that are custom-tailored to their needs without being dependent on the application provider.
Brokerpaks can generate their own documentation, alleviating the need to distribute diffs or caveats with official documentation.
Because brokerpaks have a fine-grained scope and are distributed with their own version of Terraform and providers, backwards-compatibility is less of a concern.
Because services are written with built-in test-cases, they can be used to evaluate the effects of upgrading Terraform versions, provider versions, and services.

## Best practices

### Brokerpak guidelines

Aim to keep your brokerpaks small and focused around a core idea.

It may be beneficial to divide your services into brokerpaks based on any of the following factors:

 * The users of the service, e.g. organizational unit.
 * The stability of the backing service (alpha, beta, GA).
 * The subject matter experts that work on the services, e.g. networking vs database.
 
#### Brokerpak lifecycle example

The GCP Service Broker will split its brokerpaks into three sets:

* The `preview` brokerpak will contain upcoming services. It's expected that you install the GA brokerpak, so we can freely move services from preview to GA as needed.
* The `current` brokerpak will contain the full list of services.
* The `unmaintained` brokerpaks will each contain exactly one service that we no longer support. This is so you can install exactly as many as needed and take over maintenance of any you need.

As services evolve, support can naturally pass to those who still need legacy technologies. This is a pattern you can follow in your organization too.

### Naming guidelines

Names _should_ be CLI friendly, meaning that they are alphanumeric, lower-case, and separated by dashes (-).

Service names _should_ begin with your organization and if necessary the cloud platform they're based on. To avoid collisions, you can also include the department name. For example, if your company was "Widgets Inc.":

| Name | Description |
| ----:|:----------- |
| `google-sql` | **Bad**, doesn't include your company name so it might conflict with official releases. |
| `google-widgets-sql` | **Bad**, your company name should come first. |
| `widgets-sql` | **Good** |
| `widgets-aws-sql` | **Good**, indicates the cloud platform as well as the service. |
| `widgets-acctg-sql` | **Good**, indicates that the service is maintained by/for the accounting department. |
| `legacy-widgets-sql` | **Good**, the legacy keyword comes first to indicate the service is deprecated. |

### Service guidelines

When you're creating a new service for the broker you're designing for three separate sets of people:

* The users, developers who will use your service to provision things they work with day-to-day.
* The operators, the people who are responsible for approving services and plans for developers to use.
* Yourself, the person who has to maintain the service, strike the right balance of power between the operators and users, and make sure the new plans/services work as intended.

The following sections contain guidelines to help you out.

#### Deciding what to include

Services don't need to map one-to-one with cloud products, and probably shouldn't.
Instead, services should be focused around particular workflows, allowing you to get a single, useful, task done.
Service plans allow you to scale that up or down.

For example, Google CloudSQL contains options for high availability, PostgreSQL and MySQL servers, managing on-prem servers, and read-only replication architectures.
These features all exist for different audiences, and a generic service trying to fit all the use-cases won't give a good experience to the users, operators, or maintainers.

If you find yourself wishing you could selectively enable or disable variables based on flags, it's a sign you should break down your code into another service.
For example, a Cloud Storage bucket can be configured to have a retention policy, a public-facing URL, and/or push file-change updates to a Pub/Sub queue.
It would be a good idea to break those features into multiple distinct services:

* One for hosting a static website with settings for URL, index/error pages, and CNAME.
* Another for general storage that has retention policies.
* A third that also provisions a Pub/Sub queue and acts as a staging area for data.

Breaking things down like this makes it easier to figure out what variables you need to expose, what risks they entail and what kind of plans you'll want:

* The static site plans could be simple, maybe containing different domain names and regions.
* The archive bucket plans could be for different retention policies and object durability.
* The staging bucket plans could include options for setting up alerting and the queue at the same time as the bucket is created.

Each cloud service you expose will have a plethora of tunable parameters to choose from.
Ideally, you should expose enough to be useful to developers and secure, but few enough that your service has a well-defined use-case.
You can always add more parameters later, but you can never get rid of one.

#### Deciding where to include things

Each parameter can either be set by the operator when they define plans (or in your plans that the operators enable for users) or by the user.

In general, properties which have monetary costs or affect the security of the platform should be put in the plan and properties affecting the behaviour of the resource should be defined by the user.

In our static site bucket example, the operator would create plans for different domain names (security) and bucket locations/durabilities (pricing), and the developer would get to set the parameters for the default index/error pages and maybe hostname. A full CNAME would be calculated from the hostname and domain name combination. It isn't clear who would get control over the Pub/Sub endpoint. On one hand, the developers might need it to update a search engine index, but on the other, the operator might conduct ongoing security audits.

#### Deciding on sensible defaults

The GCP Service Broker operates under the model that the users are benign but fallible.
Sensible defaults are secure and work well in the average use-case.
This highly depends on your target audience.

For example, a Pub/Sub instance with one-to-many semantics might default to a read-only role, assuming the default consumer is just going to be a worker node whereas a Pub/Sub instance with many-to-many semantics might default to a read/write role even if some consumers want to be read-only.

#### Deciding on what your default plans will be

If you've gotten to this point, you should have a clear understanding of what your service is trying to accomplish, who the users are, and what variables are configurable in your plans.
It can be tempting to include every permutation of the variables for plans.
However, less is more.
Operators need to look at each plan, decide if it fits a distinct use-case, budget and security model then make it available to individual teams.
A few plans that hit key use-cases are much easier to grok.

Let's go back to our archival storage use-case. Instead of creating plans for every availability tier and zone, we'd create plans for these criteria:

* Companies hosting archives in the US
* Companies that do not want their data in the US
* Teams that need buckets they control for non-prod environments

We'd end up with something the following:

    (US | EU | Asia) x (high availability + legally mandated retention policy | standard availability + no retention policy) = 6 plans


### Service life cycle

Each service has two interdependent life cycles: the **definition life cycle** and the **API life cycle**.

#### Definition life cycle

The **definition life cycle** reflects the state of your plugin.
It can be in one of three states, represented by `tags` on the service definition:

* `preview` - The service may have some outstanding issues, or lack documentation, but is ready for savvy users.
* (no tag) - The service is ready to be used by all users.
* `unmaintained` - The service should not be used by any users except those that already rely on it and will have no future developments.
* `eol` - End of life. The service may operate at a reduced capacity (e.g. blocking new provisioning or forcing service upgrades) due to changes in the upstream service.

#### API life cycle

The **API life cycle** reflects the state of the backing Google Cloud services your plugin depends on.
Here the published [launch stages](https://cloud.google.com/terms/launch-stages) that Google follows, and we replicate

* `beta` - There are no SLA or technical support obligations in a Beta release, and charges may be waived in some cases. Products will be complete from a feature perspective, but may have some open outstanding issues. Beta releases are suitable for limited production use cases.
* (no tag) - GA features are open to all developers and are considered stable and fully qualified for production use.
* `deprecated` - Deprecated features are scheduled to be shut down and removed.

The **API life cycle** tag MUST be set to the least supported launch stage of any of its components.
For example, if your plugin uses a deprecated API and two beta APIs, the tag would be `deprecated`.
If your plugin uses three GA APIs and a beta API then the tag would be `beta`.

  deprecated < beta < ga

NOTE: Alpha and Early Access plugins WILL NOT be included in official releases of the broker.

#### Operating with life cycles

Breaking down life cycles into distinct sets helps operators decide what amount of risk they're willing to take on.
For example, an operator might be willing to allow an unmaintained plugin if the underlying services were GA.
Alternatively, an operator might not want to enable `deprecated` plugins on a new install even if they're maintained.

## Developing and Testing a Brokerpak

See [brokerpak specificaton](./brokerpak-specification.md) and [brokerpak dissection](./brokerpak-dissection.md) for details on authoring a brokerpak.

This section will outline the development tools and lifecycle for iterating on the development of a new brokerpak (or modifying an existing brokerpak.)

### Broker docker image

There is a docker image with the broker binary that can be used to avoid having to compile the broker binary locally.

```bash
docker pull cfplatformeng/csb
```

### Building a Brokerpak

To create a brokerpak from a manifest and service yaml files:

```bash
export BROKERPAK_SRC_DIR=<absolute path to broker source directory>
docker run --rm -v ${BROKERPAK_SRC_DIR}:/brokerpak -w /brokerpak cfplatformeng/csb pak build
```

If the broker builds successfully, the result will be *.brokerpak* file in the brokerpak source directory.

### Running Examples to test a Brokerpak

If the *examples* section of the brokerpak is not empty, it is possible (and advisable) to use the examples to drive a provision, bind, unbind, and deprovision cycle for each example against a locally running broker.

> For example purposes, this is the AWS broker, so AWS credentials are provided through environment variables. See [AWS brokerpak readme](../aws-brokerpak/README.md).

In the brokerpak repo root, run:
```bash
csb run-examples --all
```

Or when using Docker:
```bash
docker run --rm -v ${BROKERPAK_SRC_DIR}:/brokerpak -w /brokerpak \
-p 8080:8080 \
-e "SECURITY_USER_NAME=csb-un" \
-e "SECURITY_USER_PASSWORD=csb-pw" \
-e AWS_ACCESS_KEY_ID \
-e AWS_SECRET_ACCESS_KEY \
cfplatformeng/csb run-examples --all
```

If this completes successfully, it means all the examples in the brokerpak successfully completed a provision, bind, unbind and deprovision lifecycle. 