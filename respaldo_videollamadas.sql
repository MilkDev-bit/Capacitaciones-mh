--
-- PostgreSQL database dump
--

\restrict ze94enHb2YbfGnNbqaJewIVRFSav2GMZjswpSBQ6YffNSlfBT3uXhJJQakV6mzX

-- Dumped from database version 18.4 (Debian 18.4-1.pgdg13+1)
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: instructor_schedules; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.instructor_schedules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    instructor_id uuid NOT NULL,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    status character varying(20) DEFAULT 'available'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.instructor_schedules OWNER TO postgres;

--
-- Name: videocall_tickets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.videocall_tickets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    capacitacion_id uuid NOT NULL,
    licencia_id uuid,
    codigo character varying(50) NOT NULL,
    in_use_by_user_id uuid,
    is_valid boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    schedule_id uuid,
    owner_id uuid
);


ALTER TABLE public.videocall_tickets OWNER TO postgres;

--
-- Data for Name: instructor_schedules; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.instructor_schedules (id, instructor_id, start_time, end_time, status, created_at) FROM stdin;
8c71dd1d-50a6-4c87-b437-e4132bc9c5e1	dae2e422-3fc0-408f-a769-c19befd0f373	2026-07-05 23:42:00+00	2026-07-06 00:42:00+00	available	2026-06-29 23:42:37.736142+00
ce13f1a0-3a93-42f4-a894-89941cba25f6	dae2e422-3fc0-408f-a769-c19befd0f373	2026-06-30 16:26:13.549+00	2026-06-30 17:26:13.549+00	available	2026-06-30 16:26:53.473167+00
b019b9e1-81f7-40e6-be0e-995e634b19a6	2a394991-668a-4ae6-92fe-5680cf0ca317	2026-06-30 16:27:36.237+00	2026-06-30 17:27:36.237+00	booked	2026-06-30 16:27:48.312806+00
3c62f4d2-9872-42c1-a43a-e1e8af46db4f	2a394991-668a-4ae6-92fe-5680cf0ca317	2026-06-30 21:49:31.573+00	2026-06-30 22:49:31.573+00	booked	2026-06-30 21:49:44.998685+00
\.


--
-- Data for Name: videocall_tickets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.videocall_tickets (id, capacitacion_id, licencia_id, codigo, in_use_by_user_id, is_valid, created_at, schedule_id, owner_id) FROM stdin;
d9f923af-1287-4da1-8f5e-2298331994a0	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-23a038c2	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
316f43b9-781d-46a5-9ca6-17a5498a699b	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-ba8d224b	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
2eea41d4-030a-49da-a871-898fae7e6a8c	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-d30532f3	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
89402da9-1436-4cd6-9cf2-086738caa280	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-4ba07ed8	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
2f892ee3-3364-4f09-aa32-c6c1dac9ef5d	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-0abbb954	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
8bbbe52d-001b-442e-b90d-bd250e1e0537	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-31113692	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
05a31db3-5d34-44a3-b779-f6d1529452a8	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-f4d0649c	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
e83aec36-663e-4728-a109-187ee3448879	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-61c4a8cc	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
31a48238-665a-4b7b-be44-002e7edb6e0c	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-1cd0e178	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
5728e4e3-c465-4e12-bde3-16301086b26a	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-c525e1ec	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
f0b4a1f9-8797-43eb-bcb5-8dd41712fcdb	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-3ff07ea1	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
b0329c6a-db30-494a-9dab-f9e29009e24f	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-5f306992	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
5cddf4ef-df69-4788-9e1f-a25f964668fa	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-15f2a640	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
8c7b07b3-420e-4afc-b1c2-d98f3eebd1c6	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-7f973a1c	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
fbc81650-1af3-4723-85a9-709845141550	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-dbaea79c	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
e8b46690-0738-4375-9fb0-a6f04d386601	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-1f05c76f	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
953c00c9-d421-4386-badf-9dc4e516fb43	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-12d0bbc2	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
c3c667d4-4bd0-4889-a917-61370d82cf4e	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-da9927db	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
c64e9ffe-a889-4f0e-85b8-9ed406c8b1f1	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-b7869cf6	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
a671c8d1-bb89-4d00-bf6d-1fb6d204e5b6	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-8ad8d6e8	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
60a70504-fab1-41ad-ac9d-645d72745c9b	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-cb72a2aa	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
485ffe2b-b688-441e-a0ee-fb7bd3584bb6	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-6277e456	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
760d772d-3bd1-4203-b5a5-67e2798cfab8	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-da40f3df	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
79b0a477-01f4-4cd4-83d8-5256d34c51fd	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-fe88080d	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
13a46b92-bbd1-4ef9-96bf-022ee482320a	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-6bd2d1f1	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
c0b30c81-aefd-45df-a0b1-f3d358293779	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	cb04b6a8-7690-4b23-8bf9-4aa1bd05fef7	VC-dee0280a	\N	t	2026-06-30 21:51:30.882872+00	\N	\N
f9859dbb-65d7-4ea8-92a4-7669fc643d9c	ef3bbae2-d06b-4cfc-a4f0-5b3bbbcf38fe	ce8f5b18-0b3f-4231-ade3-37efd20eea2e	VC-8c2dd14a	\N	t	2026-06-30 21:23:24.642305+00	\N	\N
\.


--
-- Name: instructor_schedules instructor_schedules_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.instructor_schedules
    ADD CONSTRAINT instructor_schedules_pkey PRIMARY KEY (id);


--
-- Name: videocall_tickets videocall_tickets_codigo_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.videocall_tickets
    ADD CONSTRAINT videocall_tickets_codigo_key UNIQUE (codigo);


--
-- Name: videocall_tickets videocall_tickets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.videocall_tickets
    ADD CONSTRAINT videocall_tickets_pkey PRIMARY KEY (id);


--
-- Name: videocall_tickets videocall_tickets_capacitacion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.videocall_tickets
    ADD CONSTRAINT videocall_tickets_capacitacion_id_fkey FOREIGN KEY (capacitacion_id) REFERENCES public.capacitaciones(id) ON DELETE CASCADE;


--
-- Name: videocall_tickets videocall_tickets_licencia_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.videocall_tickets
    ADD CONSTRAINT videocall_tickets_licencia_id_fkey FOREIGN KEY (licencia_id) REFERENCES public.curso_licencias(id) ON DELETE CASCADE;


--
-- Name: videocall_tickets videocall_tickets_schedule_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.videocall_tickets
    ADD CONSTRAINT videocall_tickets_schedule_id_fkey FOREIGN KEY (schedule_id) REFERENCES public.instructor_schedules(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict ze94enHb2YbfGnNbqaJewIVRFSav2GMZjswpSBQ6YffNSlfBT3uXhJJQakV6mzX

