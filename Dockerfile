ARG PROJECTNAME
ARG OUTPUTDIR

FROM golang:1.12 AS builder
ARG PROJECTNAME
WORKDIR /${PROJECTNAME}
COPY . .
RUN make build

FROM scratch
ARG PROJECTNAME
ARG OUTPUTDIR
COPY --from=builder /${PROJECTNAME}/${OUTPUTDIR}/${PROJECTNAME} /${PROJECTNAME}/${OUTPUTDIR}/${PROJECTNAME}
WORKDIR /${PROJECTNAME}/${OUTPUTDIR}/